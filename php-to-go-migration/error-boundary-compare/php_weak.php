<?php

function findUser(int $userId, string $email): array
{
    return [
        'user_id_type' => gettype($userId),
        'user_id' => $userId,
        'email_type' => gettype($email),
        'email' => $email,
    ];
}

$payload = json_decode('{"user_id":"42","email":123}', true);

try {
    $result = findUser($payload['user_id'], $payload['email']);
    echo json_encode($result, JSON_UNESCAPED_UNICODE | JSON_PRETTY_PRINT) . PHP_EOL;
} catch (Throwable $e) {
    echo get_class($e) . ': ' . $e->getMessage() . PHP_EOL;
}
