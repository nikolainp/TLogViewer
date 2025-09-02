 CREATE TABLE details (
  title             text PRIMARY KEY,
  version           text NOT NULL,
  processingSize    integer NOT NULL,
  processingSpeed   real NOT NULL,
  processingTime    integer NOT NULL,
  firstEventTime    datetime NOT NULL,
  lastEventTime     datetime NOT NULL
);

CREATE TABLE processes (
  name            text,
  catalog         text,
  process         text,
  processID       integer,
  processType     text,
  pid             text,
  port            text,
  UID             text,
  serverName      text,
  ip              text,
  firstEventTime  datetime,
  lastEventTime   datetime
);
