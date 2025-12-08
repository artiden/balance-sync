<?php

namespace App\EventMessages;

class BalanceUpdated extends EventMessage
{
    protected string $eventType = 'balance-updated';

    public function __construct(
        protected int $userId,
        protected int $walletId,
        protected int $walletBalance,
        protected string $updatedAt,
    ) {}
}
