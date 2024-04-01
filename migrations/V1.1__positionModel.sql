CREATE TABLE trading.positions (
    operation_id uuid PRIMARY KEY,
    user_id uuid,
    symbol varchar(10),
    open_price decimal,
    close_price decimal,
    created_at timestamp WITH time zone DEFAULT NOW(),
    long boolean,
);

CREATE FUNCTION capture_func() 
RETURNS trigger AS 
$$ 
BEGIN 
    IF (TG_OP = 'INSERT') THEN
        RAISE NOTICE '%', 'NOTIFY on INSERT';
        PERFORM pg_notify('positionOpen', '{"operation_id":"' || NEW.operation_id::text || '","open_price":"' || NEW.open_price::text || '","long":' || NEW.buy::text || '"}'); 
        RETURN NEW;
    ELSE
        RAISE NOTICE '%', 'NOTIFY on UPDATE';
        PERFORM pg_notify('positionClose', '{"operation_id":"' || NEW.operation_id::text || '","close_price":"' || NEW.close_price::text || '"}'); 
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER mytrigger BEFORE INSERT OR UPDATE
    ON  trading.positions FOR EACH ROW EXECUTE PROCEDURE capture_func();