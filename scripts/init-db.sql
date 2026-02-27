-- Initialize database for iacctl
-- This script runs when PostgreSQL container starts for the first time

-- Create extensions if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create additional indexes for better performance
-- These will be created by migrations, but we can add them here for initial setup

-- Grant permissions to the iacctl_user
-- Note: This is already handled by PostgreSQL's default behavior when creating the database
-- but we include it here for completeness

-- Log initialization
DO $$
BEGIN
    RAISE NOTICE 'iacctl database initialized successfully';
END $$;
