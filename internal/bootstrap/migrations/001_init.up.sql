CREATE TABLE IF NOT EXISTS public.token_transfers (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    tx_hash TEXT NOT NULL,
    log_index INTEGER NOT NULL,
    block_number BIGINT NOT NULL,
    token_address TEXT NOT NULL,
    from_address TEXT NOT NULL,
    to_address TEXT NOT NULL,
    value NUMERIC(78,0) NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    CONSTRAINT token_transfers_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_block_number 
    ON public.token_transfers USING BTREE (block_number DESC);

CREATE INDEX IF NOT EXISTS idx_from_address 
    ON public.token_transfers USING BTREE (from_address);

CREATE INDEX IF NOT EXISTS idx_to_address 
    ON public.token_transfers USING BTREE (to_address);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_transfer 
    ON public.token_transfers USING BTREE (tx_hash, log_index);