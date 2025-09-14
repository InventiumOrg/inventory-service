CREATE TABLE "inventory" (
  "id" bigserial PRIMARY KEY, 
  "name" varchar NOT NULL,
  "unit" varchar NOT NULL,
  "quantity" int NOT NULL,
  "measure" varchar NOT NULL,
  "category" varchar NOT NULL,
  "location" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);