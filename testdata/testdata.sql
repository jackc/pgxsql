create table person (
  id bigint primary key,
  name text not null
);

insert into person (id, name) values
  (1, 'Adam'),
  (2, 'Eve'),
  (3, 'Cain'),
  (4, 'Abel');
