package relay

import (
	"bytes"
	"encoding/json"
)

type Inventory []InventoryItem

func (inv *Inventory) Less(i, j int) bool {
	ii := (*inv)[i].Info
	ij := (*inv)[j].Info

	return ii.Value/ii.Weight < ij.Value/ij.Weight
}

func (inv *Inventory) Len() int {
	return len(*inv)
}

func (inv *Inventory) Swap(i, j int) {
	tmp := (*inv)[i]
	(*inv)[i] = (*inv)[j]
	(*inv)[j] = tmp
}

type InventoryItem struct {
	Name  string `json:"text"`
	Count int    `json:"count"`
	Info  Info   `json:"itemCardInfoList"`
}

type Info struct {
	Value  float32
	Weight float32
	Damage [6]float32
}

//UnmarshalJSON implements json.Unmarshaler
func (i *Info) UnmarshalJSON(bs []byte) error {
	var v []struct {
		Value, DamageRating, DiffRating, Difference float32
		ValueType, DamageType                       int
		Text                                        string
	}

	err := json.NewDecoder(bytes.NewReader(bs)).Decode(&v)
	if err != nil {
		return err
	}

	for _, val := range v {
		switch val.Text {
		case "$wt":
			i.Weight = val.Value
		case "$val":
			i.Value = val.Value
		case "$dr":
			i.Damage[val.DamageType-1] = val.Value
		}
	}

	return nil
}
