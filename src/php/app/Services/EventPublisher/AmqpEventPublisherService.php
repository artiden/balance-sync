<?php

namespace App\Services\EventPublisher;

use App\Contracts\Service\EventPublisherServiceInterface;
use Illuminate\Support\Facades\Queue;
use JsonSerializable;

class AmqpEventPublisherService implements EventPublisherServiceInterface
{
    public function __construct(
        protected string $connection,
        protected string $queue
    ) {}

    public function publish(JsonSerializable $payload): void
    {
        Queue::connection($this->connection)
            ->pushRaw(
                \json_encode($payload->jsonSerialize()),
                $this->queue
            );
    }
}
