package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Favourite holds the schema definition for the Favourite entity.
type Favourite struct {
	ent.Schema
}

// Fields of the Favourite.
func (Favourite) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").Default(time.Now),
	}
}

// Edges of the Favourite.
func (Favourite) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().
			Required(),
		edge.To("room", Room.Type).
			Unique().
			Required(),
	}
}

// Indexes of the Favourite.
func (Favourite) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("user", "room").Unique(),
	}
}
