package GraphQLdemo

import (
	"encoding/json"
	"fmt"
	"regexp"

	model1 "github.com/Go-GraphQL-Group/SW-Crawler/model"
	"github.com/boltdb/bolt"
)

const peopleBucket = "People"
const filmsBucket = "Film"
const planetsBucket = "Planet"
const speciesBucket = "Specie"
const starshipsBucket = "Starship"
const vehiclesBucket = "Vehicle"

// 正则替换
const origin = "http://localhost:8080/query/+[a-zA-Z_]+/"

const replace = ""

const preURL = "http://localhost:8080/query/"

func convertPeople(people1 *model1.People) *People {
	people2 := &People{}
	people2.ID = people1.ID
	people2.Name = people1.Name
	people2.BirthYear = &people1.Birth_year
	people2.EyeColor = &people1.Eye_color
	people2.Gender = &people1.Gender
	people2.HairColor = &people1.Hair_color
	people2.Height = &people1.Heigth
	people2.Mass = &people1.Mass
	people2.SkinColor = &people1.Skin_color
	return people2
}

func convertFilm(film1 *model1.Film) *Film {
	film2 := &Film{}
	film2.ID = film1.ID
	film2.Title = film1.Title
	film2.EpisodeID = &film1.Episode_id
	film2.OpeningCrawl = &film1.Opening_crawl
	film2.Director = &film1.Director
	film2.Producer = &film1.Producer
	film2.ReleaseDate = &film1.Release_data
	return film2
}

func GetPeopleByID(ID string, db *bolt.DB) (error, *People) {
	var err error
	if db == nil {
		db, err = bolt.Open("./data/data.db", 0600, nil)
		checkErr(err)
		defer db.Close()
	}
	people1 := &People{}

	people := &model1.People{}
	err = db.View(func(tx *bolt.Tx) error {
		peoBuck := tx.Bucket([]byte(peopleBucket))
		v := peoBuck.Get([]byte(ID))

		if v == nil {
			return err
		}

		// 正则替换
		re, _ := regexp.Compile(origin)
		rep := re.ReplaceAllString(string(v), replace)
		err = json.Unmarshal([]byte(rep), people)
		checkErr(err)
		people1 = convertPeople(people)

		// films
		for _, it := range people.Films {
			it = it[0 : len(it)-1]
			filmBuck := tx.Bucket([]byte(filmsBucket))
			film := filmBuck.Get([]byte(it))
			film1 := &model1.Film{}
			err = json.Unmarshal([]byte(film), film1)
			checkErr(err)
			people1.Films = append(people1.Films, convertFilm(film1))
		}

		people.Url = preURL + "people/" + people.Url
		return nil
	})
	return err, people1
}

// err
func checkErr(err error) {
	if err != nil {
		fmt.Println("Error occur: ", err)
		// os.Exit(1)
	}
}
