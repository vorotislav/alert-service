CREATE TABLE public.metrics (
                                "name" text NOT NULL PRIMARY KEY,
                                "type" text NOT NULL,
                                delta bigint NULL,
                                value double precision NULL
);
