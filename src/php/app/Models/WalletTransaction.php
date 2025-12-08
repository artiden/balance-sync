<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

class WalletTransaction extends Model
{
        const TYPE_AUTO = 0;
    const TYPE_DEPOSIT = 1;
    const TYPE_WITHDRAW = 2;

    protected $fillable = [
        'user_id',
        'amount',
        'type',
    ];

    protected $casts = [
        'type' => 'integer',
    ];

    public function wallet(): BelongsTo
    {
        return $this->belongsTo(Wallet::class);
    }

    public function user(): BelongsTo
    {
        return $this->belongsTo(User::class);
    }

    public function getTypeReadableAttribute(): string
    {
        return match($this->type) {
            self::TYPE_DEPOSIT => 'deposit',
            self::TYPE_WITHDRAW => 'withdraw',
            default => 'auto',
        };
    }
}
