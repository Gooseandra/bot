create table exercises(
    id serial primary key,
    exercise varchar(20) NOT NULL,
    chatID int not null
);

create table test(
  value int,
  exercise int references exercises(id) not null
);

create table attributes(
   exercise int references exercises(id) not null,
    "name" varchar(20)

);

create table results(
    execId int references exercises(id) not null ,
    trainId int references state(id) not null,
    attribute varchar(20),
    value int
);

create table state(
    id serial primary key,
    date date,
    startTime time,
    endTime time,
    comment varchar(500),
    userId int not null
);

create table forjson(
    id serial primary key,
    data varchar(200)
)
