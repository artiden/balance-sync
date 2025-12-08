<?php

namespace App\Providers;

use App\Contracts\Service\EventPublisherServiceInterface;
use App\Services\EventPublisher\AmqpEventPublisherService;
use Illuminate\Contracts\Foundation\Application;
use Illuminate\Support\ServiceProvider;

class EventPublisherServiceProvider extends ServiceProvider
{
    /**
     * Register services.
     */
    public function register(): void
    {
        // We could use singleton, synce it haven't any state and so on...
        $this->app->singleton(EventPublisherServiceInterface::class, function(Application $app){
            return new AmqpEventPublisherService(
                connection: config('walletevents.connection'),
                queue: config('walletevents.queue')
            );
        });
    }

    /**
     * Bootstrap services.
     */
    public function boot(): void
    {
        //
    }
}
