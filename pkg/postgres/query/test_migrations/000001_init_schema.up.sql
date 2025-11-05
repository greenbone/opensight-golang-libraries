-- note: as this is merely for testing and not applied to a persistent DB,
-- we can just edit this sql file, rather than actually adding migrations

CREATE TABLE test_table (
    "id" INT PRIMARY KEY,
    "string" TEXT,
    "integer" INT,
    "float" FLOAT4,
    "boolean" BOOLEAN,
    "date_time" TIMESTAMPTZ
);
