CREATE TABLE public.shorts
(
    short      STRING    NOT NULL,
    url        STRING    NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT "primary" PRIMARY KEY (short ASC),
    FAMILY "primary" (short, url, created_at)
);
