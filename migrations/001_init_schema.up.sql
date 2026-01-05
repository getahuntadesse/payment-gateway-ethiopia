-- Ethiopian Payment Gateway Database Schema
-- Supports ETB (Ethiopian Birr) and USD payments

-- Create payments table
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL CHECK (currency IN ('ETB', 'USD')),
    reference VARCHAR(100) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' 
        CHECK (status IN ('PENDING', 'SUCCESS', 'FAILED')),
    description VARCHAR(500),
    customer_name VARCHAR(200),
    bank_code VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payments_reference ON payments(reference);
CREATE INDEX IF NOT EXISTS idx_payments_bank_code ON payments(bank_code);
CREATE INDEX IF NOT EXISTS idx_payments_currency ON payments(currency);

-- Create function to auto-update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for updated_at
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
CREATE TRIGGER update_payments_updated_at
    BEFORE UPDATE ON payments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert sample Ethiopian payments
INSERT INTO payments (id, amount, currency, reference, status, description, customer_name, bank_code) VALUES
    ('11111111-1111-1111-1111-111111111111', 5000.00, 'ETB', 'CBE-2023-000001', 'SUCCESS', 
     'á‹¨á‰¡áŠ“ áˆ�áˆ­á‰µ áŠ­á��á‹« (Coffee product payment)', 'Yesuf Ahimed', 'CBE'),
    ('22222222-2222-2222-2222-222222222222', 250.75, 'USD', 'AWASH-2023-000002', 'FAILED', 
     'Software subscription', 'Tech Solutions PLC', 'AWASH'),
    ('33333333-3333-3333-3333-333333333333', 15000.50, 'ETB', 'DASHEN-2023-000003', 'PENDING', 
     'áˆˆáŠ•áŒ�á‹µ á‰¤á‰µ áŠªáˆ«á‹­ (Office rent)', 'Love', 'DASHEN'),
    ('44444444-4444-4444-4444-444444444444', 1000.00, 'ETB', 'CBE-2023-000004', 'SUCCESS', 
     'Education fee', 'marry Bob', 'CBE'),
    ('55555555-5555-5555-5555-555555555555', 500.00, 'USD', 'ABYSSINIA-2023-000005', 'SUCCESS', 
     'Export payment for leather products', 'Addis Leather Co.', 'ABYSSINIA')
ON CONFLICT (reference) DO NOTHING;

-- Create view for Ethiopian payment statistics
CREATE OR REPLACE VIEW payment_statistics AS
SELECT 
    COUNT(*) as total_payments,
    SUM(CASE WHEN currency = 'ETB' THEN amount ELSE 0 END) as total_etb,
    SUM(CASE WHEN currency = 'USD' THEN amount ELSE 0 END) as total_usd,
    COUNT(CASE WHEN status = 'SUCCESS' THEN 1 END) as successful_payments,
    COUNT(CASE WHEN status = 'FAILED' THEN 1 END) as failed_payments,
    COUNT(CASE WHEN status = 'PENDING' THEN 1 END) as pending_payments,
    ROUND(AVG(CASE WHEN currency = 'ETB' THEN amount END), 2) as avg_etb_amount,
    ROUND(AVG(CASE WHEN currency = 'USD' THEN amount END), 2) as avg_usd_amount
FROM payments;

-- Comment on tables (for documentation)
COMMENT ON TABLE payments IS 'Stores Ethiopian payment transactions in ETB and USD';
COMMENT ON COLUMN payments.currency IS 'Currency code: ETB (Ethiopian Birr) or USD (US Dollar)';
COMMENT ON COLUMN payments.bank_code IS 'Ethiopian bank code: CBE, AWASH, DASHEN, etc.';
COMMENT ON COLUMN payments.customer_name IS 'Name of Ethiopian customer (in Amharic or English)';
COMMENT ON COLUMN payments.description IS 'Payment description in Amharic or English';