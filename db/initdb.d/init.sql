CREATE TABLE companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(15) UNIQUE NOT NULL,
    description TEXT,
    amount_of_employees INT NOT NULL,
    registered BOOLEAN NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('Corporations', 'NonProfit', 'Cooperative', 'Sole Proprietorship'))
);

CREATE TABLE admin_users (
    username VARCHAR(255) PRIMARY KEY,
    hashed_password VARCHAR(255) NOT NULL
);
INSERT INTO admin_users (username, hashed_password) VALUES ('admin', '$2b$12$B.j0F3qX39NVj.gvv.VLCOld/.1FwMUiOsi80l/aUwyfn.ac2wVZa');
