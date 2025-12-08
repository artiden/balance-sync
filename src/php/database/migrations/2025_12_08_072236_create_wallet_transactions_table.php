<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Run the migrations.
     */
    public function up(): void
    {
        Schema::create('wallet_transactions', function (Blueprint $table) {
            $table->id();
            // To be able get all user transactions without join (Discussed de-normalization)
            $table->foreignId('user_id')
                ->constrained()
                ->cascadeOnDelete();
            $table->foreignId('wallet_id')
                ->constrained()
                ->cascadeOnDelete();
            $table->bigInteger('amount');
            $table->tinyInteger('type')
                ->default(0);
            $table->index('user_id');
            $table->index('wallet_id');
            $table->index('type');
            $table->timestamps();
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists('wallet_transactions');
    }
};
