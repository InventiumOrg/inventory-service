CREATE TABLE "inventory" (
  "id" bigserial PRIMARY KEY, 
  "name" varchar NOT NULL,
  "quantity" int NOT NULL,
  "category" char NOT NULL,
  "located" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);