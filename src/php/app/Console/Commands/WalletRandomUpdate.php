<?php

namespace App\Console\Commands;

use App\Models\User;
use App\Models\WalletTransaction;
use App\Contracts\Service\WalletServiceInterface;
use Illuminate\Console\Command;

class WalletRandomUpdate extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'wallet:update-random';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = 'Updates random users balances';

    public function __construct(
        protected WalletServiceInterface $walletService
    ) {
        parent::__construct();
    }

    /**
     * Execute the console command.
     */
    public function handle()
    {
        $userCount = \random_int(5, 50);

        $users = User::inRandomOrder()
            ->limit($userCount)
            ->with('wallet')
            ->get();

        foreach ($users as $user) {
            // To be sure user have a wallet
            if (\is_null($user->wallet)) {
                continue;
            }

            $delta = \random_int(-1000, 1000);

            $this->walletService->updateBalance(
                $user->wallet,
                $delta,
                WalletTransaction::TYPE_AUTO
            );

            $this->info(\sprintf(
                'User %d: Balance updated by: %d',
                $user->id,
                $delta
            ));
        }
    }
}
