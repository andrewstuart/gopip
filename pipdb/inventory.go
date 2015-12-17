package pipdb

import (
	"bytes"
	"encoding/json"
	"math"
)

type Inventory struct {
	I []InventoryItem
	V func(InventoryItem) float32
}

func (inv *Inventory) Less(i, j int) bool {
	vi := inv.V(inv.I[i])
	vj := inv.V(inv.I[j])

	return math.IsNaN(float64(vj)) || vi < vj
}

func (inv *Inventory) Len() int {
	return len(inv.I)
}

func (inv *Inventory) Swap(i, j int) {
	inv.I[i], inv.I[j] = inv.I[j], inv.I[i]
}

type InventoryItem struct {
	HandleID    int
	StackID     []int
	CanFavorite bool
	IsLegendary bool
	Name        string `json:"text"`
	Count       int    `json:"count"`
	Info        Info   `json:"itemCardInfoList"`
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
