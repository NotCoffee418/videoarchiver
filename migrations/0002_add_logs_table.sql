-- +up
CREATE TABLE IF NOT EXISTS "logs" (
    "id" INTEGER NOT NULL UNIQUE,
    "verbosity" INTEGER NOT NULL,
    "timestamp" BIGINT NOT NULL DEFAULT (strftime('%s', 'now')),
    "message" TEXT NOT NULL,
    PRIMARY KEY("id")
);

CREATE INDEX "logs_index_0"
ON "logs" ("verbosity", "timestamp");

-- +down
DROP TABLE IF EXISTS "logs";