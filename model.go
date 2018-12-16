package GraphQLdemo

// A People resource is an individual person or character within the Star Wars universe.
type People struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	BirthYear *string `json:"birth_year"`
	EyeColor  *string `json:"eye_color"`
	Gender    *string `json:"gender"`
	HairColor *string `json:"hair_color"`
	Height    *string `json:"height"`
	Mass      *string `json:"mass"`
	SkinColor *string `json:"skin_color"`
	Films     []*Film `json:"films"`
}
