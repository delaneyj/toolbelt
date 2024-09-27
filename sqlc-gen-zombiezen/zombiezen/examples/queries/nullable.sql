-- name: GetNullableTestTableMyBool :many
SELECT
    *
FROM
    nullableTestTable
WHERE
    myBool = @myBool;