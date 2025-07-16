-- name: CreateChirps :one
insert into chirps(id, created_at, updated_at, body, user_id)
values (gen_random_uuid(), now(), now(), $1, $2)
returning *;

-- name: GetChirps :many
select * from chirps
order by created_at;

-- name: GetChirpsByAuthor :many
select * from chirps
where user_id = $1
order by created_at;

-- name: GetChirp :one
select * from chirps
where id = $1;

-- name: DeleteChirp :execrows
delete from chirps
where id = $1 and user_id = $2;
