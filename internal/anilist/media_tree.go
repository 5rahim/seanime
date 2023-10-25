package anilist

import (
	"context"
	"github.com/seanime-app/seanime-server/internal/limiter"
	"github.com/seanime-app/seanime-server/internal/result"
	"sync"
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
	wg := sync.WaitGroup{}

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

	for _, _edge := range edges {

		// Edge is not a sequel or prequel, skip
		if *_edge.RelationType != MediaRelationSequel && *_edge.RelationType != MediaRelationPrequel {
			continue
		}

		if *_edge.GetNode().Status == MediaStatusNotYetReleased {
			continue
		}

		// Edge is TV, TV_SHORT, SPECIAL, MOVIE, OVA, ONA
		if _edge.IsBroadRelationFormat() && !tree.Has(_edge.GetNode().ID) {
			wg.Add(1)

			// A prequel edge will only fetch its prequels
			// A sequel edge will only fetch its sequels
			_edgeRel := rel
			if rel == "all" && *_edge.RelationType == MediaRelationPrequel {
				_edgeRel = FetchMediaTreePrequels
			}
			if rel == "all" && *_edge.RelationType == MediaRelationSequel {
				_edgeRel = FetchMediaTreeSequels
			}

			go func(edge *BaseMedia_Relations_Edges, edgeRel *string) {

				// Find the edge in the tree
				cacheV, ok := cache.Get(edge.GetNode().ID)

				// Continue with the media is in the tree or fetch it
				edgeBaseMedia := cacheV
				cont := false
				if !ok {
					// Wait for the rate limiter
					rateLimiter.Wait()
					println("cache MISS", edge.GetNode().ID)
					res, err := anilistClient.BaseMediaByID(context.Background(), &edge.GetNode().ID)
					if err == nil {
						edgeBaseMedia = res.GetMedia()
						cache.Set(edgeBaseMedia.ID, edgeBaseMedia)
						cont = true
					}

				} else {
					println("cache HIT", edge.GetNode().ID)
					cont = true
				}

				if cont {
					_ = edgeBaseMedia.FetchMediaTree(*edgeRel, anilistClient, rateLimiter, tree, cache)
				}

				wg.Done()

			}(_edge, &_edgeRel)

		}
	}

	// Wait until all the edges and their edges are fetched
	wg.Wait()

	return nil

}
