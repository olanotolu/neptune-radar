-- Expand church_jurisdictions.jurisdiction_type to support non-Catholic
-- denominations: Episcopal dioceses, UMC annual conferences, Jewish federations.
ALTER TABLE church_jurisdictions DROP CONSTRAINT IF EXISTS church_jurisdictions_jurisdiction_type_check;
ALTER TABLE church_jurisdictions ADD CONSTRAINT church_jurisdictions_jurisdiction_type_check
  CHECK (jurisdiction_type IN ('diocese','archdiocese','annual_conference','missionary_conference','federation'));
