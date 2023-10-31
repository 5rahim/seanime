package anilist

import (
	"context"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime-server/internal/limiter"
	"github.com/seanime-app/seanime-server/internal/result"
)

type BaseMediaRelationTree struct {
	*result.Map[int, *BaseMedia]
}

// NewBaseMediaRelationTree returns a new result.Map[int, *BaseMedia].
// It is used to store the results of FetchMediaTree or FetchMediaTree calls.
func NewBaseMediaRelationTree() *BaseMediaRelationTree {
	return &BaseMediaRelationTree{result.NewResultMap[int, *BaseMedia]()}
}

type FetchMediaTreeRelation = string

const (
	FetchMediaTreeSequels  FetchMediaTreeRelation = "sequels"
	FetchMediaTreePrequels FetchMediaTreeRelation = "prequels"
	FetchMediaTreeAll      FetchMediaTreeRelation = "all"
)

// FetchMediaTree populates the BaseMediaRelationTree with the given media's sequels and prequels.
// It also takes a BaseMediaCache to store the fetched media in and avoid duplicate fetches.
// It also takes a limiter.Limiter to limit the number of requests made to the AniList API.
func (m *BaseMedia) FetchMediaTree(rel FetchMediaTreeRelation, anilistClient *Client, rateLimiter *limiter.Limiter, tree *BaseMediaRelationTree, cache *BaseMediaCache) error {

	// If the media is in the result cache, skip
	if tree.Has(m.ID) {
		cache.Set(m.ID, m)
		return nil
	} else {
		cache.Set(m.ID, m)
		// Add the media to the result tree
		tree.Set(m.ID, m)
	}

	// If the media does not have relations, skip
	if m.Relations == nil {
		return nil
	}

	edges := m.GetRelations().GetEdges()

	// Filter out edges that are not sequels or prequels, not released yet, not specific format, or already in the tree
	edges = lo.Filter(edges, func(_edge *BaseMedia_Relations_Edges, _ int) bool {
		if *_edge.RelationType != MediaRelationSequel && *_edge.RelationType != MediaRelationPrequel {
			return false
		}
		if *_edge.GetNode().Status == MediaStatusNotYetReleased {
			return false
		}
		if !_edge.IsBroadRelationFormat() || tree.Has(_edge.GetNode().ID) {
			return false
		}
		return true
	})

	// If there are no edges left, skip
	if len(edges) == 0 {
		return nil
	}

	// Create a channel to wait for all goroutines to finish
	doneCh := make(chan struct{})

	// For each edge, fetch the media and add it to the tree
	lop.ForEach(edges, func(edge *BaseMedia_Relations_Edges, _ int) {
		edgeRel := rel
		// If the relation is "all", but the edge is a prequel, fetch the edge's prequels
		if rel == "all" && *edge.RelationType == MediaRelationPrequel {
			edgeRel = FetchMediaTreePrequels
		}
		// If the relation is "all", but the edge is a sequel, fetch the edge's sequels
		if rel == "all" && *edge.RelationType == MediaRelationSequel {
			edgeRel = FetchMediaTreeSequels
		}

		// Find the edge in the cache
		cacheV, ok := cache.Get(edge.GetNode().ID)

		edgeBaseMedia := cacheV
		cont := false

		// If the edge is not in the cache, fetch it
		if !ok {
			// Wait for the rate limiter
			rateLimiter.Wait()
			res, err := anilistClient.BaseMediaByID(context.Background(), &edge.GetNode().ID)
			if err == nil {
				edgeBaseMedia = res.GetMedia()
				cache.Set(edgeBaseMedia.ID, edgeBaseMedia)
				cont = true
			}

		} else {
			cont = true
		}

		if cont {
			// Fetch the edge's relations
			edgeBaseMedia.FetchMediaTree(edgeRel, anilistClient, rateLimiter, tree, cache)
		}

	})

	go func() {
		close(doneCh)
	}()

	for {
		select {
		case <-doneCh:
			return nil
		default:
		}
	}

}

//----------------------------------------------------------------------------------------------------------------------
