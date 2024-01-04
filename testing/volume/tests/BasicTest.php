<?php declare(strict_types=1);

namespace Tests;

use PHPUnit\Framework\TestCase;

final class BasicTest extends TestCase
{
    public function testBasicExample(): void
    {
        dump(\Foo\TestEnum::ZERO);
        $this->assertSame("a", "a");
    }
}