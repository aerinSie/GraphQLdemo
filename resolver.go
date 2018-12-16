package GraphQLdemo

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/boltdb/bolt"
)

type Resolver struct{}

func (r *Resolver) People() PeopleResolver {
	return &peopleResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type peopleResolver struct{ *Resolver }

func (r *peopleResolver) FilmConnection(ctx context.Context, obj *People, first *int, after *string) (FilmConnection, error) {
	from := -1
	if after != nil {
		b, err := base64.StdEncoding.DecodeString(*after)
		if err != nil {
			return FilmConnection{}, err
		}
		i, err := strconv.Atoi(strings.TrimPrefix(string(b), "cursor"))
		if err != nil {
			return FilmConnection{}, err
		}
		from = i
	}
	index := -1
	count := 0
	hasPreviousPage := false
	hasNextPage := true
	// 获取edges
	edges := []FilmEdge{}
	for i, film := range obj.Films {
		if film.ID == strconv.Itoa(from) {
			index = i
			break
		}
	}
	if index > 0 {
		hasPreviousPage = true
	}
	for i := index + 1; i < len(obj.Films); i++ {
		edges = append(edges, FilmEdge{
			Node:   obj.Films[i],
			Cursor: encodeCursor(obj.Films[i].ID),
		})
		count++
		if count >= *first {
			break
		}
	}
	if count < *first {
		hasNextPage = false
	}
	if count == 0 {
		return FilmConnection{}, nil
	}
	// 获取pageInfo
	pageInfo := PageInfo{
		HasPreviousPage: hasPreviousPage,
		HasNextPage:     hasNextPage,
		StartCursor:     encodeCursor(edges[0].Node.ID),
		EndCursor:       encodeCursor(edges[count-1].Node.ID),
	}

	return FilmConnection{
		PageInfo:   pageInfo,
		Edges:      edges,
		TotalCount: count,
	}, nil
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) People(ctx context.Context, id string) (*People, error) {
	err, people := GetPeopleByID(id, nil)
	checkErr(err)
	return people, err
}

func (r *queryResolver) Peoples(ctx context.Context, first *int, after *string) (PeopleConnection, error) {
	from := -1
	if after != nil {
		b, err := base64.StdEncoding.DecodeString(*after)
		if err != nil {
			return PeopleConnection{}, err
		}
		i, err := strconv.Atoi(strings.TrimPrefix(string(b), "cursor"))
		if err != nil {
			return PeopleConnection{}, err
		}
		from = i
	}
	count := 0
	startID := ""
	hasPreviousPage := true
	hasNextPage := true
	// 获取edges
	edges := []PeopleEdge{}
	db, err := bolt.Open("./data/data.db", 0600, nil)
	checkErr(err)
	defer db.Close()
	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(peopleBucket)).Cursor()

		// 判断是否还有前向页
		k, v := c.First()
		if from == -1 || strconv.Itoa(from) == string(k) {
			startID = string(k)
			hasPreviousPage = false
		}

		if from == -1 {
			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				_, people := GetPeopleByID(string(k), db)
				edges = append(edges, PeopleEdge{
					Node:   people,
					Cursor: encodeCursor(string(k)),
				})
				count++
				if count == *first {
					break
				}
			}
		} else {
			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				if strconv.Itoa(from) == string(k) {
					k, _ = c.Next()
					startID = string(k)
				}
				if startID != "" {
					_, people := GetPeopleByID(string(k), db)
					edges = append(edges, PeopleEdge{
						Node:   people,
						Cursor: encodeCursor(string(k)),
					})
					count++
					if count == *first {
						break
					}
				}
			}
		}

		k, v = c.Next()
		if k == nil && v == nil {
			hasNextPage = false
		}
		return nil
	})
	if count == 0 {
		return PeopleConnection{}, nil
	}
	// 获取pageInfo
	pageInfo := PageInfo{
		HasPreviousPage: hasPreviousPage,
		HasNextPage:     hasNextPage,
		StartCursor:     encodeCursor(startID),
		EndCursor:       encodeCursor(edges[count-1].Node.ID),
	}

	return PeopleConnection{
		PageInfo:   pageInfo,
		Edges:      edges,
		TotalCount: count,
	}, nil
}

func encodeCursor(k string) string {
	i, _ := strconv.Atoi(k)
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("cursor%d", i)))
}
