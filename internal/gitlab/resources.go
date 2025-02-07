package gitlabClient

import (
	"context"
	"fmt"
	"glpatEye/internal/metrics"
	"glpatEye/pkg/common"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
)

const (
	ResourceProject = "project"
	ResourceGroup   = "group"
)

// Function is getting list IDs, name and web-url for projects via graphQL and create objectMap,
// where key is project ID and value is a struct with some fields.
func ProcessProjects(ctx context.Context, client *GitlabClient, responseObjectsSize, poolSize int) {
	var counterProjects int32
	s := &client.Response.Data.Projects
	s.PageInfo.HasNextPage = true
	s.PageInfo.EndCursor = ""
	for s.PageInfo.HasNextPage {
		s.Nodes = nil
		err := client.GetIdsResource(ctx, responseObjectsSize, s.PageInfo.EndCursor, "projects")
		if err != nil {
			log.Printf("error fetching projects, %s", err)
			break
		}
		s.PageInfo = client.Response.Data.Projects.PageInfo
		projectMap := buildObjectMap(s.Nodes, true)
		cs := CheckSignature{
			Ctx:          ctx,
			Client:       client,
			ObjectMap:    projectMap,
			ResourceType: ResourceProject,
			Counter:      &counterProjects,
		}
		processTokens(cs, poolSize)
	}
	log.Printf("Total scanned projects: %d", counterProjects)
}

// It's the same thing, but main entity is group.
func ProcessGroups(ctx context.Context, client *GitlabClient, responseObjectsSize, poolSize int) {
	var counterGroups int32
	s := &client.Response.Data.Groups
	s.PageInfo.HasNextPage = true
	s.PageInfo.EndCursor = ""
	for s.PageInfo.HasNextPage {
		s.Nodes = nil
		err := client.GetIdsResource(ctx, responseObjectsSize, s.PageInfo.EndCursor, "groups")
		if err != nil {
			log.Printf("error fetching groups, %s", err)
			break
		}
		s.PageInfo = client.Response.Data.Groups.PageInfo
		groupMap := buildObjectMap(s.Nodes, false)
		cs := CheckSignature{
			Ctx:          ctx,
			Client:       client,
			ObjectMap:    groupMap,
			ResourceType: ResourceGroup,
			Counter:      &counterGroups,
		}
		processTokens(cs, poolSize)
	}
	log.Printf("Total scanned groups: %d", counterGroups)
}

func buildObjectMap(nodes []ProjectNode, project bool) map[string]ProjectNode {
	objectMap := make(map[string]ProjectNode)
	for _, id := range nodes {
		numericID := common.ExtractNumericID(id.ID)
		if !project {
			objectMap[numericID] = ProjectNode{
				Name:   id.Name,
				WebUrl: id.WebUrl,
			}
		} else {
			objectMap[numericID] = ProjectNode{
				Name:          id.Name,
				HttpUrlToRepo: id.HttpUrlToRepo,
			}
		}
		// log.Printf("added to project map: %s\t%s", numericID, id.Name)
	}
	return objectMap
}

// The main function for checking all tokens at projects and groups.
func processTokens(s CheckSignature, poolSize int) {
	resultCh := make(chan []AccessToken, len(s.ObjectMap))
	errorCh := make(chan error, len(s.ObjectMap))
	idCh := make(chan string, len(s.ObjectMap))

	for id := range s.ObjectMap {
		// log.Printf("Sending ID to channel: %s", id)
		idCh <- id
	}
	close(idCh)
	wg := sync.WaitGroup{}
	var workers int = poolSize

	cs := CheckSignature{
		Ctx:           s.Ctx,
		Client:        s.Client,
		IDs:           idCh,
		ObjectMap:     s.ObjectMap,
		ResourceType:  s.ResourceType,
		ResultChannel: resultCh,
		ErrorChannel:  errorCh,
		Counter:       s.Counter,
	}

	// log.Printf("Start time processing: %s\n", time.Now().Local())
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go checkTokens(cs, &wg)
	}
	wg.Wait()
	// log.Printf("End time: %s\n", time.Now().Local())
	close(resultCh)
	close(errorCh)
}

func CheckMasterToken(ctx context.Context, client *GitlabClient) {
	masterToken, err := client.SelfCheckMasterToken(ctx, client.Token)
	if err != nil {
		log.Printf("Failed to check master token, %s\n", err)
		return
	}
	metric := metrics.MetricLabels{
		Name:        masterToken.Name,
		ProjectName: "root",
		UrlToRepo:   fmt.Sprintf("%s/admin", client.BaseURL),
		Id:          strconv.Itoa(masterToken.ID),
		LastUsed:    masterToken.LastUsed,
		Root:        strconv.FormatBool(true),
		DaysExpire:  masterToken.DaysExpire,
	}
	metric.UpdateMetric()
	log.Printf("Result checking gitlab master token: %v", masterToken)
}

func checkTokens(s CheckSignature, wg *sync.WaitGroup) {
	defer wg.Done()
	var matchingTokens = make([]AccessToken, 0)
	for id := range s.IDs {
		result, err := s.Client.CheckAccessTokens(s.Ctx, id, s.ObjectMap, s.ResourceType)
		if err != nil {
			log.Printf("Failed to check token for project, %s\n", err)
			s.ErrorChannel <- err
			break
		}
		s.ResultChannel <- result
		matchingTokens = append(matchingTokens, result...)
		if len(matchingTokens) > 0 {
			log.Printf("Matching tokens with pattern: %v\n", matchingTokens)
		}
		atomic.AddInt32(s.Counter, 1)
	}
}
