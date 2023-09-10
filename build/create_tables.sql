CREATE DATABASE TransactionalDB;
ALTER ROLE admin SET client_encoding TO 'utf8';
GRANT ALL PRIVILEGES ON DATABASE TransactionalDB TO admin;

CREATE TABLE clients (
    id SERIAL PRIMARY KEY,
    account_number NUMERIC(20,0) UNIQUE NOT NULL,
    card_number NUMERIC(16,0) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE currencies (
    id SERIAL PRIMARY KEY,
    alfa_code VARCHAR(3) UNIQUE NOT NULL,
    number_code NUMERIC(3,0) UNIQUE NOT NULL,
    name VARCHAR(127) NOT NULL
);

CREATE TABLE balances (
    id SERIAL PRIMARY KEY,
    client_id INT REFERENCES clients(id) NOT NULL,
    currency_id INT REFERENCES currencies(id) NOT NULL,
    actual_balance MONEY NOT NULL,
    frozen_balance MONEY NOT NULL
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    client_id INT REFERENCES clients(id),
    status VARCHAR(7) NOT NULL,
    currency_id INT REFERENCES currencies(id) NOT NULL,
    amount MONEY NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Creating a new transaction and adding the amount to the client's balance
CREATE OR REPLACE FUNCTION create_invoice(
    p_client_account NUMERIC,
    p_currency VARCHAR(3),
    p_amount MONEY
) RETURNS INTEGER AS $$
DECLARE
    v_client_id INT;
    v_currency_id INT;
    v_transaction_id INT;
BEGIN
    -- Verification of the user's existence by account number
    SELECT id INTO v_client_id FROM clients WHERE account_number = p_client_account OR card_number = p_client_account;
    IF v_client_id IS NULL THEN
        RAISE EXCEPTION 'Error: User not found';
    END IF;

    -- Checking the existence of an account in the specified currency
    SELECT id INTO v_currency_id FROM currencies WHERE alfa_code = p_currency;
    IF v_currency_id IS NULL THEN
        RAISE EXCEPTION 'Error: Currency not found';
    END IF;

    -- Creating a new transaction
    INSERT INTO transactions (client_id, status, currency_id, amount)
    VALUES (v_client_id, 'Created', v_currency_id, p_amount)
    RETURNING id INTO v_transaction_id;

    -- Check if the transaction was created successfully
    IF v_transaction_id IS NULL THEN
        RAISE EXCEPTION 'Error: Error creating transaction';
    END IF;

    -- Add the amount to the client's balance
    UPDATE balances SET actual_balance = actual_balance + p_amount
    WHERE client_id = v_client_id AND currency_id = v_currency_id;

    -- Check if the balance was updated successfully
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Error: Error updating balance';
    END IF;

    -- Commit the transaction and change the status to Success
    UPDATE transactions SET status = 'Success' WHERE id = v_transaction_id;

    -- Return the ID of the created transaction
    RETURN v_transaction_id;

EXCEPTION
    WHEN OTHERS THEN
        -- An error occurred, change the status to Error
        IF v_transaction_id IS NOT NULL THEN
            UPDATE transactions SET status = 'Error' WHERE id = v_transaction_id;
        END IF;

        -- Raise the original error
        RAISE EXCEPTION '%', SQLERRM;
END;
$$ LANGUAGE plpgsql;


-- Withdraw funds from the client's account
CREATE OR REPLACE FUNCTION withdraw_funds(
    p_client_account NUMERIC,
    p_currency VARCHAR(3),
    p_amount MONEY
) RETURNS INTEGER AS $$
DECLARE
    v_client_id INT;
    v_currency_id INT;
    v_transaction_id INT;
BEGIN
    -- Verify the user's existence by account number or card number
    SELECT id INTO v_client_id FROM clients WHERE account_number = p_client_account OR card_number = p_client_account;
    IF v_client_id IS NULL THEN
        RAISE EXCEPTION 'Error: User not found';
    END IF;

    -- Check the existence of an account in the specified currency
    SELECT id INTO v_currency_id FROM currencies WHERE alfa_code = p_currency;
    IF v_currency_id IS NULL THEN
        RAISE EXCEPTION 'Error: Currency not found';
    END IF;

    -- Create a new transaction for withdrawal
    INSERT INTO transactions (client_id, status, currency_id, amount)
    VALUES (v_client_id, 'Created', v_currency_id, p_amount * -1)
    RETURNING id INTO v_transaction_id;

    -- Check if the transaction was created successfully
    IF v_transaction_id IS NULL THEN
        RAISE EXCEPTION 'Error: Error creating a new transaction';
    END IF;

    -- Check if the user has sufficient balance for withdrawal
    IF p_amount > (SELECT actual_balance FROM balances WHERE client_id = v_client_id AND currency_id = v_currency_id) THEN
        RAISE EXCEPTION 'Error: Insufficient balance for withdrawal';
    END IF;

    -- Freeze the money in the account
    UPDATE balances SET actual_balance = actual_balance - p_amount
    WHERE client_id = v_client_id AND currency_id = v_currency_id;

    -- Check if the balance was updated successfully
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Error: Error updating balance';
    END IF;

    -- Update the frozen balance
    UPDATE balances SET frozen_balance = frozen_balance + p_amount
    WHERE client_id = v_client_id AND currency_id = v_currency_id;

    -- Check if the balance was updated successfully
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Error: Error updating balance';
    END IF;

    -- Unfreeze the money in the account
    UPDATE balances SET frozen_balance = frozen_balance - p_amount
    WHERE client_id = v_client_id AND currency_id = v_currency_id;

    -- Check if the balance was updated successfully
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Error: Error updating balance';
    END IF;

    -- Update the transaction status to Success
    UPDATE transactions SET status = 'Success' WHERE id = v_transaction_id;

    -- Check if the balance was updated successfully
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Error: Error updating transaction';
    END IF;

    -- Return the ID of the created transaction
    RETURN v_transaction_id;
    
EXCEPTION
    WHEN OTHERS THEN
        -- An error occurred, change the status to Error
        IF v_transaction_id IS NOT NULL THEN
            UPDATE transactions SET status = 'Error' WHERE id = v_transaction_id;
        END IF;

        -- Raise the original error
        RAISE EXCEPTION '%', SQLERRM;
END;
$$ LANGUAGE plpgsql;

-- Get client balances in all currencies
CREATE OR REPLACE FUNCTION get_client_balances(
    p_client_account NUMERIC
) RETURNS TABLE (
    currency_code VARCHAR(3),
    actual_balance MONEY,
    frozen_balance MONEY
) AS $$
DECLARE
    v_client_id INT;
BEGIN
    -- Verify the user's existence by account number or card number
    SELECT id INTO v_client_id FROM clients WHERE account_number = p_client_account OR card_number = p_client_account;
    IF v_client_id IS NULL THEN
        -- User not found
        RETURN QUERY SELECT NULL::VARCHAR, NULL::MONEY, NULL::MONEY;
    END IF;

    RETURN QUERY
    SELECT c.alfa_code, b.actual_balance, b.frozen_balance
    FROM balances b
    JOIN clients cl ON b.client_id = cl.id
    JOIN currencies c ON b.currency_id = c.id
    WHERE cl.id = v_client_id;
END;
$$ LANGUAGE plpgsql;


-- The ISO 4217 standard is used for currency codes

INSERT INTO currencies (alfa_code, number_code, name)
VALUES
    ('AED', 784, 'United Arab Emirates Dirham'),
    ('AFN', 971, 'Afghan Afghani'),
    ('ALL', 008, 'Albanian Lek'),
    ('AMD', 051, 'Armenian Dram'),
    ('ANG', 532, 'Netherlands Antillean Guilder'),
    ('AOA', 973, 'Angolan Kwanza'),
    ('ARS', 032, 'Argentine Peso'),
    ('AUD', 036, 'Australian Dollar'),
    ('AWG', 533, 'Aruban Florin'),
    ('AZN', 944, 'Azerbaijani Manat'),
    ('BAM', 977, 'Bosnia-Herzegovina Convertible Mark'),
    ('BBD', 052, 'Barbadian Dollar'),
    ('BDT', 050, 'Bangladeshi Taka'),
    ('BGN', 975, 'Bulgarian Lev'),
    ('BHD', 048, 'Bahraini Dinar'),
    ('BIF', 108, 'Burundian Franc'),
    ('BMD', 060, 'Bermudian Dollar'),
    ('BND', 096, 'Brunei Dollar'),
    ('BOB', 068, 'Bolivian Boliviano'),
    ('BRL', 986, 'Brazilian Real'),
    ('BSD', 044, 'Bahamian Dollar'),
    ('BTN', 064, 'Bhutanese Ngultrum'),
    ('BWP', 072, 'Botswanan Pula'),
    ('BYN', 933, 'Belarusian Ruble'),
    ('BZD', 084, 'Belize Dollar'),
    ('CAD', 124, 'Canadian Dollar'),
    ('CDF', 976, 'Congolese Franc'),
    ('CHF', 756, 'Swiss Franc'),
    ('CLP', 152, 'Chilean Peso'),
    ('CNY', 156, 'Chinese Yuan'),
    ('COP', 170, 'Colombian Peso'),
    ('CRC', 188, 'Costa Rican Colón'),
    ('CUP', 192, 'Cuban Peso'),
    ('CVE', 132, 'Cape Verdean Escudo'),
    ('CZK', 203, 'Czech Republic Koruna'),
    ('DJF', 262, 'Djiboutian Franc'),
    ('DKK', 208, 'Danish Krone'),
    ('DOP', 214, 'Dominican Peso'),
    ('DZD', 012, 'Algerian Dinar'),
    ('EGP', 818, 'Egyptian Pound'),
    ('ERN', 232, 'Eritrean Nakfa'),
    ('ETB', 230, 'Ethiopian Birr'),
    ('EUR', 978, 'Euro'),
    ('FJD', 242, 'Fijian Dollar'),
    ('FKP', 238, 'Falkland Islands Pound'),
    ('GBP', 826, 'British Pound Sterling'),
    ('GEL', 981, 'Georgian Lari'),
    ('GGP', 831, 'Guernsey Pound'),
    ('GHS', 936, 'Ghanaian Cedi'),
    ('GIP', 292, 'Gibraltar Pound'),
    ('GMD', 270, 'Gambian Dalasi'),
    ('GNF', 324, 'Guinean Franc'),
    ('GTQ', 320, 'Guatemalan Quetzal'),
    ('GYD', 328, 'Guyanaese Dollar'),
    ('HKD', 344, 'Hong Kong Dollar'),
    ('HNL', 340, 'Honduran Lempira'),
    ('HRK', 191, 'Croatian Kuna'),
    ('HTG', 332, 'Haitian Gourde'),
    ('HUF', 348, 'Hungarian Forint'),
    ('IDR', 360, 'Indonesian Rupiah'),
    ('ILS', 376, 'Israeli New Sheqel'),
    ('IMP', 833, 'Isle of Man Pound'),
    ('INR', 356, 'Indian Rupee'),
    ('IQD', 368, 'Iraqi Dinar'),
    ('IRR', 364, 'Iranian Rial'),
    ('ISK', 352, 'Icelandic Króna'),
    ('JEP', 832, 'Jersey Pound'),
    ('JMD', 388, 'Jamaican Dollar'),
    ('JOD', 400, 'Jordanian Dinar'),
    ('JPY', 392, 'Japanese Yen'),
    ('KES', 404, 'Kenyan Shilling'),
    ('KGS', 417, 'Kyrgystani Som'),
    ('KHR', 116, 'Cambodian Riel'),
    ('KID', 296, 'Kiribati Dollar'),
    ('KMF', 174, 'Comorian Franc'),
    ('KPW', 408, 'North Korean Won'),
    ('KRW', 410, 'South Korean Won'),
    ('KWD', 414, 'Kuwaiti Dinar'),
    ('KYD', 136, 'Cayman Islands Dollar'),
    ('KZT', 398, 'Kazakhstani Tenge'),
    ('LAK', 418, 'Laotian Kip'),
    ('LBP', 422, 'Lebanese Pound'),
    ('LKR', 144, 'Sri Lankan Rupee'),
    ('LRD', 430, 'Liberian Dollar'),
    ('LSL', 426, 'Lesotho Loti'),
    ('LYD', 434, 'Libyan Dinar'),
    ('MAD', 504, 'Moroccan Dirham'),
    ('MDL', 498, 'Moldovan Leu'),
    ('MGA', 969, 'Malagasy Ariary'),
    ('MKD', 807, 'Macedonian Denar'),
    ('MMK', 104, 'Myanma Kyat'),
    ('MNT', 496, 'Mongolian Tugrik'),
    ('MOP', 446, 'Macanese Pataca'),
    ('MRU', 929, 'Mauritanian Ouguiya'),
    ('MUR', 480, 'Mauritian Rupee'),
    ('MVR', 462, 'Maldivian Rufiyaa'),
    ('MWK', 454, 'Malawian Kwacha'),
    ('MXN', 484, 'Mexican Peso'),
    ('MYR', 458, 'Malaysian Ringgit'),
    ('MZN', 943, 'Mozambican Metical'),
    ('NAD', 516, 'Namibian Dollar'),
    ('NGN', 566, 'Nigerian Naira'),
    ('NIO', 558, 'Nicaraguan Córdoba'),
    ('NOK', 578, 'Norwegian Krone'),
    ('NPR', 524, 'Nepalese Rupee'),
    ('NZD', 554, 'New Zealand Dollar'),
    ('OMR', 512, 'Omani Rial'),
    ('PAB', 590, 'Panamanian Balboa'),
    ('PEN', 604, 'Peruvian Nuevo Sol'),
    ('PGK', 598, 'Papua New Guinean Kina'),
    ('PHP', 608, 'Philippine Peso'),
    ('PKR', 586, 'Pakistani Rupee'),
    ('PLN', 985, 'Polish Zloty'),
    ('PYG', 600, 'Paraguayan Guarani'),
    ('QAR', 634, 'Qatari Rial'),
    ('RON', 946, 'Romanian Leu'),
    ('RSD', 941, 'Serbian Dinar'),
    ('RUB', 643, 'Russian Ruble'),
    ('RWF', 646, 'Rwandan Franc'),
    ('SAR', 682, 'Saudi Riyal'),
    ('SBD', 090, 'Solomon Islands Dollar'),
    ('SCR', 690, 'Seychellois Rupee'),
    ('SDG', 938, 'Sudanese Pound'),
    ('SEK', 752, 'Swedish Krona'),
    ('SGD', 702, 'Singapore Dollar'),
    ('SHP', 654, 'Saint Helena Pound'),
    ('SLL', 694, 'Sierra Leonean Leone'),
    ('SOS', 706, 'Somali Shilling'),
    ('SRD', 968, 'Surinamese Dollar'),
    ('SSP', 728, 'South Sudanese Pound'),
    ('STN', 930, 'São Tomé and Príncipe Dobra'),
    ('SYP', 760, 'Syrian Pound'),
    ('SZL', 748, 'Swazi Lilangeni'),
    ('THB', 764, 'Thai Baht'),
    ('TJS', 972, 'Tajikistani Somoni'),
    ('TMT', 934, 'Turkmenistani Manat'),
    ('TND', 788, 'Tunisian Dinar'),
    ('TOP', 776, 'Tongan Paʻanga'),
    ('TRY', 949, 'Turkish Lira'),
    ('TTD', 780, 'Trinidad and Tobago Dollar'),
    ('TWD', 901, 'New Taiwan Dollar'),
    ('TZS', 834, 'Tanzanian Shilling'),
    ('UAH', 980, 'Ukrainian Hryvnia'),
    ('UGX', 800, 'Ugandan Shilling'),
    ('USD', 840, 'United States Dollar'),
    ('UYU', 858, 'Uruguayan Peso'),
    ('UZS', 860, 'Uzbekistan Som'),
    ('VES', 928, 'Venezuelan Bolívar'),
    ('VND', 704, 'Vietnamese Dong'),
    ('VUV', 548, 'Vanuatu Vatu'),
    ('WST', 882, 'Samoan Tala'),
    ('XAF', 950, 'Central African CFA Franc'),
    ('XCD', 951, 'East Caribbean Dollar'),
    ('XDR', 960, 'Special Drawing Rights'),
    ('XOF', 952, 'West African CFA Franc'),
    ('XPF', 953, 'CFP Franc'),
    ('YER', 886, 'Yemeni Rial'),
    ('ZAR', 710, 'South African Rand'),
    ('ZMW', 967, 'Zambian Kwacha'),
    ('ZWL', 932, 'Zimbabwean Dollar');

-- Create 5 users with default account pairs
INSERT INTO clients (account_number, card_number) VALUES
(12345678901234567890, 1234567890123456),
(23456789012345678901, 2345678901234567),
(34567890123456789012, 3456789012345678),
(45678901234567890123, 4567890123456789),
(56789012345678901234, 5678901234567890);

-- Create balance records for each user and default currency
INSERT INTO balances (client_id, currency_id, actual_balance, frozen_balance) VALUES
(1, 1, 1000.00, 0.00),
(1, 2, 500.00, 0.00),
(2, 1, 750.00, 0.00),
(2, 3, 250.00, 0.00),
(3, 1, 1200.00, 0.00),
(3, 2, 300.00, 0.00),
(4, 2, 100.00, 0.00),
(5, 3, 800.00, 0.00);
