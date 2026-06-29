SELECT EXISTS (
    SELECT 1    // 1 is like _ in for loop of python, don't care about the value.
    FROM users
    WHERE username = $1
        AND is_deleted = FALSE
);     // return 'true' or 'false'