<?php

namespace App\Services\Wallet;

use App\Contracts\Service\EventPublisherServiceInterface;
use App\Contracts\Service\WalletServiceInterface;
use App\EventMessages\BalanceUpdated;
use App\Models\Wallet;
use App\Models\WalletTransaction;
use Illuminate\Support\Facades\DB;

class WalletService implements WalletServiceInterface
{
    public function __construct(
        protected EventPublisherServiceInterface $publisher
    ) {}

    public function updateBalance(Wallet $wallet, int $amount, int $type = WalletTransaction::TYPE_AUTO): void
    {
        // As we discussed, we could implement retry, go via queue and other stuff here.
        // Also, if it's required, should throw an exception if, for example, balance after update going be <0 and so on.
        DB::transaction(function() use ($wallet, $amount, $type){
            $wallet = Wallet::where('id', $wallet->id)
                ->lockForUpdate()
                ->firstOrFail();

            // Store a transaction for balance history
            $wallet->transactions()->create([
                'user_id' => $wallet->user_id,
                'amount' => $amount,
                'type' => $type,
            ]);

            // All balances stored in cents.
            $wallet->balance += $amount;

            $wallet->save();

            $this->publisher->publish(new BalanceUpdated(
                $wallet->user_id,
                $wallet->id,
                $wallet->balance,
                // To use the same timezone. We could configure globally, but it's not necessary for MVP
                $wallet->updated_at->setTimezone('UTC')->toISOString()
            ));
        });
    }
}