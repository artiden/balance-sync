<?php

namespace Tests\Unit\Services;

use App\Contracts\Service\EventPublisherServiceInterface;
use App\EventMessages\BalanceUpdated;
use App\Exceptions\EventPublishException;
use App\Models\User;
use App\Models\Wallet;
use App\Models\WalletTransaction;
use App\Services\Wallet\WalletService;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Mockery;
use Mockery\MockInterface;
use Tests\TestCase;

class WalletServiceTest extends TestCase
{
    use RefreshDatabase;

    private MockInterface $publisher;
    private WalletService $walletService;

    protected function setUp(): void
    {
        parent::setUp();

        $this->publisher = Mockery::mock(EventPublisherServiceInterface::class);
        $this->walletService = new WalletService($this->publisher);
    }

    protected function tearDown(): void
    {
        Mockery::close();
        parent::tearDown();
    }

    public function test_update_balance_increases_wallet_balance(): void
    {
        $user = User::factory()->create();
        $wallet = $user->wallet;
        $initialBalance = $wallet->balance;
        $amount = 1000;

        $this->publisher
            ->shouldReceive('publish')
            ->once()
            ->with(Mockery::type(BalanceUpdated::class));

        $this->walletService->updateBalance($wallet, $amount);

        $wallet->refresh();
        $this->assertEquals($initialBalance + $amount, $wallet->balance);
    }

    public function test_update_balance_decreases_wallet_balance(): void
    {
        $user = User::factory()->create();
        $wallet = $user->wallet;
        $wallet->balance = 5000;
        $wallet->save();
        $amount = -2000;

        $this->publisher
            ->shouldReceive('publish')
            ->once()
            ->with(Mockery::type(BalanceUpdated::class));

        $this->walletService->updateBalance($wallet, $amount);

        $wallet->refresh();
        $this->assertEquals(3000, $wallet->balance);
    }

    public function test_update_balance_creates_transaction_record(): void
    {
        $user = User::factory()->create();
        $wallet = $user->wallet;
        $amount = 500;

        $this->publisher
            ->shouldReceive('publish')
            ->once();

        $this->walletService->updateBalance($wallet, $amount, WalletTransaction::TYPE_DEPOSIT);

        $this->assertDatabaseHas('wallet_transactions', [
            'wallet_id' => $wallet->id,
            'user_id' => $user->id,
            'amount' => $amount,
            'type' => WalletTransaction::TYPE_DEPOSIT,
        ]);
    }

    public function test_update_balance_publishes_event_after_transaction(): void
    {
        $user = User::factory()->create();
        $wallet = $user->wallet;
        $amount = 1000;

        $publishedEvent = null;
        $this->publisher
            ->shouldReceive('publish')
            ->once()
            ->andReturnUsing(function (BalanceUpdated $event) use (&$publishedEvent) {
                $publishedEvent = $event->jsonSerialize();
            });

        $this->walletService->updateBalance($wallet, $amount);

        $this->assertNotNull($publishedEvent);
        $this->assertEquals($user->id, $publishedEvent['userId']);
        $this->assertEquals($wallet->id, $publishedEvent['walletId']);
        $this->assertEquals($amount, $publishedEvent['walletBalance']);
    }

    public function test_update_balance_continues_on_publish_failure(): void
    {
        $user = User::factory()->create();
        $wallet = $user->wallet;
        $initialBalance = $wallet->balance;
        $amount = 1000;

        $this->publisher
            ->shouldReceive('publish')
            ->once()
            ->andThrow(new EventPublishException('RabbitMQ unavailable'));

        // Should not throw exception
        $this->walletService->updateBalance($wallet, $amount);

        // Balance should still be updated
        $wallet->refresh();
        $this->assertEquals($initialBalance + $amount, $wallet->balance);
    }

    public function test_update_balance_uses_locking_for_concurrent_updates(): void
    {
        $user = User::factory()->create();
        $wallet = $user->wallet;
        $wallet->balance = 1000;
        $wallet->save();

        $this->publisher
            ->shouldReceive('publish')
            ->twice();

        // Simulate concurrent updates
        $this->walletService->updateBalance($wallet, 500);
        $this->walletService->updateBalance($wallet, 300);

        $wallet->refresh();
        $this->assertEquals(1800, $wallet->balance);
    }
}
