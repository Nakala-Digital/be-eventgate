-- EVG-40: Seed Initial Roles Migration (Up)
-- Based on EVG-34 Data Entity & Relationship Draft Revision 3

INSERT INTO roles (role_name, description) VALUES
('super_admin', 'Kendali penuh platform, pengelola sistem global & reviewer event'),
('admin_panitia', 'Organizer/EO internal, mengelola Event milik sendiri'),
('staf_lapangan', 'Petugas operasional scan QR & check-in di lokasi event'),
('school_reviewer', 'Role eksplisit opsional bila approval event dipisah dari super_admin')
ON CONFLICT (role_name) DO UPDATE 
SET description = EXCLUDED.description;
