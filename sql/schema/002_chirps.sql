-- +goose UP
create table chirps (
  id uuid,
  created_at timestamp not null,
  updated_at timestamp not null,
  body text not null,
  user_id uuid not null references users(id) on delete cascade,
  primary key(id)
);

-- +goose Down
drop table chirps;
