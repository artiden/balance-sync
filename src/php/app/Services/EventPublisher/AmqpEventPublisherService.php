<?php

namespace App\Services\EventPublisher;

use App\Contracts\Service\EventPublisherServiceInterface;
use App\Exceptions\EventPublishException;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Queue;
use JsonSerializable;

class AmqpEventPublisherService implements EventPublisherServiceInterface
{
    public function __construct(
        protected string $connection,
        protected string $queue
    ) {}

    /**
     * @throws EventPublishException
     */
    public function publish(JsonSerializable $payload): void
    {
        try {
            $data = $payload->jsonSerialize();
            $json = \json_encode($data, JSON_THROW_ON_ERROR);

            Queue::connection($this->connection)
                ->pushRaw($json, $this->queue);

            Log::debug('Event published successfully', [
                'queue' => $this->queue,
                'eventId' => $data['eventId'] ?? null,
            ]);
        } catch (\JsonException $e) {
            Log::error('Failed to encode event payload', [
                'error' => $e->getMessage(),
            ]);
            throw new EventPublishException('Failed to encode event: ' . $e->getMessage(), 0, $e);
        } catch (\Throwable $e) {
            Log::error('Failed to publish event to queue', [
                'queue' => $this->queue,
                'error' => $e->getMessage(),
            ]);
            throw new EventPublishException('Failed to publish event: ' . $e->getMessage(), 0, $e);
        }
    }
}
