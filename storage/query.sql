
-- name: GetDetails :one
SELECT * FROM details
LIMIT 1;

-- name: SaveDetails :exec
INSERT INTO details (
    title, version, 
    processingSize, processingSpeed, processingTime,
    firstEventTime, lastEventTime  
) VALUES (
  ?, ?, ?, ?, ?, ?, ?
);

-- name: GetProcesses :many
SELECT * FROM processes 
WHERE name in (sqlc.slice('names'));
