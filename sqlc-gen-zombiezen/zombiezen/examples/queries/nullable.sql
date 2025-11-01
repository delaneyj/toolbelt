-- name: GetNullableTestTableMyBool :many
SELECT
    *
FROM
    nullableTestTable
WHERE
    myBool = @myBool;

-- name: DistinctAuthorNames :many
SELECT
    DISTINCT na.name AS first_name,
    nb.name AS last_name
FROM
    authors a
    INNER JOIN NAMES na ON a.first_name_id = na.id
    INNER JOIN NAMES nb ON a.last_name_id = nb.id
ORDER BY
    first_name,
    last_name;

-- name: HasAuthors :one
SELECT
    COUNT(*) > 0
FROM
    authors;

-- name: ListNullableTestTableByIDs :many
SELECT
    id,
    myBool
FROM
    nullableTestTable
WHERE
    id IN (sqlc.slice('ids'));
