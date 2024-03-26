CREATE TABLE trading.positions (
    operation_id uuid PRIMARY KEY,
    user_id uuid,
    symbol varchar(10),
    open_price decimal,
    close_price decimal,
    created_at timestamp WITH time zone DEFAULT NOW(),
    buy boolean,
    OPEN boolean
);