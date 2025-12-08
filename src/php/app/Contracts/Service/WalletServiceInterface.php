<?php

namespace App\Contracts\Service;

use App\Models\Wallet;

interface WalletServiceInterface
{
    // In the production we should implement extra logic to throw some exceptions, return expected values and so on...
    public function updateBalance(Wallet $wallet, int $amount, int $type): void;
}
