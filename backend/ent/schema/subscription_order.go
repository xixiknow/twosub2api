package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SubscriptionOrder holds the schema definition for the SubscriptionOrder entity.
type SubscriptionOrder struct {
	ent.Schema
}

func (SubscriptionOrder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "subscription_orders"},
	}
}

func (SubscriptionOrder) Fields() []ent.Field {
	return []ent.Field{
		field.String("order_no").
			MaxLen(64).
			NotEmpty().
			Unique(),
		field.String("trade_no").
			MaxLen(128).
			Optional().
			Nillable(),
		field.Int64("user_id"),
		field.Int64("group_id"),
		field.Float("amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}),
		field.Float("original_price").
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}),
		field.Float("discount_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}).
			Default(0),
		field.String("payment_method").
			MaxLen(32).
			NotEmpty(),
		field.String("status").
			MaxLen(16).
			Default("pending"),
		field.String("notify_data").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Time("created_at").
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Immutable(),
		field.Time("paid_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("expired_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("activated_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Int64("subscription_id").
			Optional().
			Nillable(),
	}
}

func (SubscriptionOrder) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("subscription_orders").
			Field("user_id").
			Unique().
			Required(),
		edge.From("group", Group.Type).
			Ref("subscription_orders").
			Field("group_id").
			Unique().
			Required(),
	}
}

func (SubscriptionOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_no").Unique(),
		index.Fields("user_id"),
		index.Fields("status"),
	}
}
