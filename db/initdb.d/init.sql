CREATE TABLE companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(15) UNIQUE NOT NULL,
    description TEXT,
    amount_of_employees INT NOT NULL,
    registered BOOLEAN NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('Corporations', 'NonProfit', 'Cooperative', 'Sole Proprietorship'))
);
