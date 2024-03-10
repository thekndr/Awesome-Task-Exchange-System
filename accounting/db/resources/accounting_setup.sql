CREATE TABLE IF NOT EXISTS accounting_workers (
       id SERIAL PRIMARY KEY,
       public_id UUID NOT NULL,
       email VARCHAR(255) UNIQUE NOT NULL
       balance INT NOT NULL DEFAULT 0
);

-- DROP TYPE IF EXISTS accounting_billing_cycle_status;
-- CREATE TYPE accounting_billing_cycle_status AS ENUM ('active', 'closed');
CREATE TABLE IF NOT EXISTS accounting_billing_cycles (
       id SERIAL PRIMARY KEY,
       created_at TIMESTAMP NOT NULL DEFAULT now(),
       -- TODO: closed_at
       status VARCHAR(10)
);

-- DROP TYPE IF EXISTS accounting_transaction_operation;
-- CREATE TYPE accounting_transaction_operation AS ENUM ('enrollment', 'withdrawal', 'payment');
CREATE TABLE IF NOT EXISTS accounting_transactions (
       id SERIAL PRIMARY KEY,
       created_at TIMESTAMP NOT NULL DEFAULT now(),
       withdrawal INT NOT NULL,
       enrolment INT NOT NULL,
       FOREIGN KEY (id) REFERENCES accounting_tasks(id),
       FOREIGN KEY (id) REFERENCES accounting_workers(id),
       FOREIGN KEY (id) REFERENCES accounting_billing_cycles(id)
);

CREATE TABLE IF NOT EXISTS accounting_payments (
       id SERIAL PRIMARY KEY,
       created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS accounting_tasks (
       id SERIAL PRIMARY KEY,
       public_id UUID NOT NULL,
       description VARCHAR(255),
       completed BOOLEAN NOT NULL,
       assignment_price INT NOT NULL,
       reward_price INT NOT NULL
);
