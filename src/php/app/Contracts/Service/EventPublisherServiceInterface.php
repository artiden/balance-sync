<?php

namespace App\Contracts\Service;

interface EventPublisherServiceInterface
{
    // In the production we should implement extra logic to throw some exceptions, return expected values and so on...
    public function publish(\JsonSerializable $payload): void;
}
