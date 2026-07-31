-- EVG-40: Seed Initial Roles Migration (Down)
-- Based on EVG-34 Data Entity & Relationship Draft Revision 3

DELETE FROM roles WHERE role_name IN ('super_admin', 'admin_panitia', 'staf_lapangan', 'school_reviewer');
