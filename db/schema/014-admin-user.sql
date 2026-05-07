insert into users (first_name, last_name, email, password_hash, user_type, image_path, is_active)
values (
    'Admin',
    'User',
    'admin@ceocw.local',
    '$2a$12$NMYYEclzXTxKTwYqZ6hkJeDBtGZ4N9FlXN/mrM0sawzAVO60pPsqi',
    'admin',
    '',
    1
)
on duplicate key update
    first_name = values(first_name),
    last_name = values(last_name),
    password_hash = values(password_hash),
    user_type = values(user_type),
    image_path = values(image_path),
    is_active = values(is_active);
