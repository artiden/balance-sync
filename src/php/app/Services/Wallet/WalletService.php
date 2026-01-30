<?php

namespace App\Services\Wallet;

use App\Contracts\Service\EventPublisherServiceInterface;
use App\Contracts\Service\WalletServiceInterface;
use App\EventMessages\BalanceUpdated;
use App\Exceptions\EventPublishException;
use App\Models\Wallet;
use App\Models\WalletTransaction;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Log;

class WalletService implements WalletServiceInterface
{
    public function __construct(
        protected EventPublisherServiceInterface $publisher
    ) {}

    public function updateBalance(Wallet $wallet, int $amount, int $type = WalletTransaction::TYPE_AUTO): void
    {
        // As we discussed, we could implement retry, go via queue and other stuff here.
        // Also, if it's required, should throw an exception if, for example, balance after update going be <0 and so on.

        // Prepare event data after transaction commits (not inside transaction)
        $eventData = null;

        DB::transaction(function() use ($wallet, $amount, $type, &$eventData){
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

            // Capture event data to publish after transaction commits
            $eventData = [
                'userId' => $wallet->user_id,
                'walletId' => $wallet->id,
                'balance' => $wallet->balance,
                // To use the same timezone. We could configure globally, but it's not necessary for MVP
                'updatedAt' => $wallet->updated_at->setTimezone('UTC')->toISOString(),
            ];
        });

        // Publish event after transaction is committed successfully
        // This ensures DB changes are persisted before event is sent
        if ($eventData !== null) {
            try {
                $this->publisher->publish(new BalanceUpdated(
                    $eventData['userId'],
                    $eventData['walletId'],
                    $eventData['balance'],
                    $eventData['updatedAt']
                ));
            } catch (EventPublishException $e) {
                // Log the failure but don't rollback the transaction
                // In production, we should implement a retry mechanism or outbox pattern
                Log::error('Failed to publish balance update event', [
                    'walletId' => $eventData['walletId'],
                    'userId' => $eventData['userId'],
                    'error' => $e->getMessage(),
                ]);
            }
        }
    }
}