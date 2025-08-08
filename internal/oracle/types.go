package oracle

type Choice struct {
	Options []string `json:"options"`
}

type Table struct {
	Name string `json:"name"`
}

type Dice struct {
	Count int `json:"count"`
	Sides int `json:"sides"`
}

type Text struct {
	Value string `json:"value"`
}
