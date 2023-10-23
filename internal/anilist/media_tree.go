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
// It is used to store the results of FetchMediaTree or FetchMediaTreeC calls.
func NewBaseMediaRelationTree() *BaseMediaRelationTree {
	return &BaseMediaRelationTree{result.NewResultMap[int, *BaseMedia]()}
}

type FetchMediaTreeRelation = string

const (
	FetchMediaTreeSequels  FetchMediaTreeRelation = "sequels"
	FetchMediaTreePrequels FetchMediaTreeRelation = "prequels"
	FetchMediaTreeAll      FetchMediaTreeRelation = "all"
)

// FetchMediaTree populates the `cache` with the given media's sequels and prequels.
func (m *BaseMedia) FetchMediaTree(rel FetchMediaTreeRelation, anilistClient *Client, rateLimiter *limiter.Limiter, tree *BaseMediaRelationTree) error {

	wg := sync.WaitGroup{}

	// If the media is in the result cache, skip
	if tree.Has(m.ID) {
		return nil
	} else {
		// Add the media to the result tree
		tree.Set(m.ID, m)
	}

	// If the media does not have relations, skip
	if m.GetRelations() == nil {
		return nil
	}

	edges := m.GetRelations().GetEdges()

	for _, _edge := range edges {

		// Edge is not a sequel or prequel, skip
		if _edge.GetRelationType().String() != MediaRelationSequel.String() && _edge.GetRelationType().String() != MediaRelationPrequel.String() {
			continue
		}

		// Edge is TV, TV_SHORT, SPECIAL, MOVIE, OVA, ONA
		if _edge.IsBroadRelationFormat() {

			wg.Add(1)

			// A prequel edge will only fetch its prequels
			// A sequel edge will only fetch its sequels
			_edgeRel := rel
			if rel == "all" && _edge.GetRelationType().String() == MediaRelationPrequel.String() {
				_edgeRel = FetchMediaTreePrequels
			}
			if rel == "all" && _edge.GetRelationType().String() == MediaRelationSequel.String() {
				_edgeRel = FetchMediaTreeSequels
			}

			go func(edge *BaseMedia_Relations_Edges, edgeRel *string) {

				defer wg.Done()

				// Find the edge in the tree
				treeV, ok := tree.Get(edge.GetNode().ID)

				// Continue with the media is in the tree or fetch it
				edgeBaseMedia := treeV
				cont := false
				if !ok {
					// Wait for the rate limiter
					rateLimiter.Wait()

					res, err := anilistClient.BaseMediaByID(context.Background(), &edge.GetNode().ID)
					if err == nil {
						edgeBaseMedia = res.GetMedia()
						cont = true
					}

				} else {
					cont = true
				}

				if cont {
					_ = edgeBaseMedia.FetchMediaTree(*edgeRel, anilistClient, rateLimiter, tree)
				}

			}(_edge, &_edgeRel)

		}
	}

	// Wait until all the edges and their edges are fetched
	wg.Wait()

	return nil

}

// FetchMediaTreeC populates the BaseMediaRelationTree with the given media's sequels and prequels.
// It also takes a BaseMediaCache to store the fetched media in and avoid duplicate fetches.
// It also takes a limiter.Limiter to limit the number of requests made to the AniList API.
func (m *BaseMedia) FetchMediaTreeC(rel FetchMediaTreeRelation, anilistClient *Client, rateLimiter *limiter.Limiter, tree *BaseMediaRelationTree, cache *BaseMediaCache) error {

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
	if m.GetRelations() == nil {
		return nil
	}

	edges := m.GetRelations().GetEdges()

	for _, _edge := range edges {

		// Edge is not a sequel or prequel, skip
		if _edge.GetRelationType().String() != MediaRelationSequel.String() && _edge.GetRelationType().String() != MediaRelationPrequel.String() {
			continue
		}

		// Edge is TV, TV_SHORT, SPECIAL, MOVIE, OVA, ONA
		if _edge.IsBroadRelationFormat() && !tree.Has(_edge.GetNode().ID) {

			wg.Add(1)

			// A prequel edge will only fetch its prequels
			// A sequel edge will only fetch its sequels
			_edgeRel := rel
			if rel == "all" && _edge.GetRelationType().String() == MediaRelationPrequel.String() {
				_edgeRel = FetchMediaTreePrequels
			}
			if rel == "all" && _edge.GetRelationType().String() == MediaRelationSequel.String() {
				_edgeRel = FetchMediaTreeSequels
			}

			go func(edge *BaseMedia_Relations_Edges, edgeRel *string) {

				defer wg.Done()

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
					_ = edgeBaseMedia.FetchMediaTreeC(*edgeRel, anilistClient, rateLimiter, tree, cache)
				}

			}(_edge, &_edgeRel)

		}
	}

	// Wait until all the edges and their edges are fetched
	wg.Wait()

	return nil

}
