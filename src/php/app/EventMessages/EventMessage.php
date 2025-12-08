<?php

namespace App\EventMessages;

use Illuminate\Support\Str;

abstract class EventMessage implements \JsonSerializable
{
    public function jsonSerialize(): array
    {
        return \array_merge(
            // To track an event in the services interested in. To be handled 1 time only.
            [
                'eventId' => (string)Str::uuid(),
            ],
            \get_object_vars($this),
        );
    }
}
