<?php

return [
    'connection' => env('BALANCE_EVENTS_CONNECTION', 'rabbitmq'),
    'queue' => env('BALANCE_EVENTS_QUEUE', 'wallet_events'),
];
