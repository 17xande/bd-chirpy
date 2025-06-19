-- +goose UP
create table users (
  id uuid,
  created_at timestamp not null,
  updated_at timestamp not null,
  email text not null unique,
  PRIMARY KEY(id)
);

-- +goose Down
drop table users
