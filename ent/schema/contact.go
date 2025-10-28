package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Contact holds the schema definition for the Contact entity.
type Contact struct {
	ent.Schema
}

// Fields of the Contact.
func (Contact) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("is_favourite").Default(false),
		field.Bool("is_blocked").Default(false),
		field.String("alias").Default(""),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Contact.
func (Contact) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("owner", User.Type).
			Unique().
			Required(),
		edge.To("contact", User.Type).
			Unique().
			Required(),
	}
}

// Indexes of the Contact.
func (Contact) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("owner", "contact").Unique(),
	}
}
