CREATE TABLE IF NOT EXISTS roles (
    name            VARCHAR(20)        PRIMARY KEY,
    description     VARCHAR(150)       NOT NULL,
    created_at      TIMESTAMPTZ        NOT NULL DEFAULT NOW()
);

INSERT INTO roles (name, description)
SELECT 'admin', 'Akses penuh terhadap seluruh data dan pengaturan'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'admin');

INSERT INTO roles (name, description)
SELECT 'staff', 'Boleh melihat data seluruh user, tetapi tidak boleh mengubah'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'staff');

INSERT INTO roles (name, description)
SELECT 'user', 'Hanya boleh mengelola datanya sendiri'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'user');

CREATE TABLE IF NOT EXISTS permissions (
    name        VARCHAR(50)     PRIMARY KEY,
    description VARCHAR(150)    NOT NULL
);

INSERT INTO permissions (name, description)
SELECT 'student:list', 'Melihat daftar seluruh student'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'student:list');
INSERT INTO permissions (name, description)
SELECT 'student:read:any', 'Melihat data student mana pun'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'student:read:any');
INSERT INTO permissions (name, description)
SELECT 'student:create', 'Membuat data student'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'student:create');
INSERT INTO permissions (name, description)
SELECT 'student:update:any', 'Mengubah data student mana pun'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'student:update:any');
INSERT INTO permissions (name, description)
SELECT 'student:delete', 'Menghapus student'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'student:delete');
INSERT INTO permissions (name, description)
SELECT 'role:assign', 'Mengubah role milik student lain'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'role:assign');

CREATE TABLE IF NOT EXISTS role_permissions (
    role_name       VARCHAR(20)     NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
    permission_name VARCHAR(50)     NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', 'student:list'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'admin' AND permission_name = 'student:list');
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', 'student:read:any'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'admin' AND permission_name = 'student:read:any');
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', 'student:create'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'admin' AND permission_name = 'student:create');
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', 'student:update:any'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'admin' AND permission_name = 'student:update:any');
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', 'student:delete'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'admin' AND permission_name = 'student:delete');
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', 'role:assign'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'admin' AND permission_name = 'role:assign');
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'staff', 'student:list'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'staff' AND permission_name = 'student:list');
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'staff', 'student:read:any'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'staff' AND permission_name = 'student:read:any');
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'staff', 'student:create'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'staff' AND permission_name = 'student:create');

UPDATE users SET role = 'user' WHERE role NOT IN (SELECT name FROM roles);

ALTER TABLE users DROP CONSTRAINT users_role_fkey;
ALTER TABLE users
    ADD CONSTRAINT users_role_fkey
    FOREIGN KEY (role) REFERENCES roles(name);

CREATE INDEX IF NOT EXISTS users_role_idx ON users (role); 
