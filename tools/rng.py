"""原版亂數產生器的參考模型（`docs/re/13`）。

對應 `wl.merged.exe` 的四支函式：

    sub_18E6B  raw()        原始產生器，五個位元組的進位鏈
    sub_18E5F  d6()         1..6
    sub_18E41  roll(n)      1..n（遮罩 ＋ 拒絕取樣）
    sub_19D86  sum_d6()     累加 n 顆 d6，16-bit 飽和
    sub_19C84  pair_d6()    2d6，逢同點續擲並累加

這是 RE 的驗證工具，不是引擎程式碼——引擎要等規格標 READY 才動工。

自我測試：
    python3 tools/rng.py
"""

from __future__ import annotations


class Rng:
    """`sub_18E6B` 的狀態：`ds:465Ch`–`ds:4660h` 五個位元組，映像裡初值全為 0。"""

    def __init__(self, state: list[int] | None = None) -> None:
        self.state = list(state) if state else [0, 0, 0, 0, 0]
        self.al = 0  # 呼叫端沿用的 AL，是回饋鏈的一部分

    def raw(self, al: int | None = None) -> int:
        """sub_18E6B：計數器加一，再把 AL 沿五個位元組做進位鏈累加。"""
        if al is not None:
            self.al = al & 0xFF
        s = self.state
        s[0] = (s[0] + 1) & 0xFF

        total = self.al + s[0]  # add al, ds:465Ch
        carry = total >> 8
        acc = total & 0xFF
        for k in range(1, 5):  # adc al, ds:465Dh … ds:4660h
            total = acc + s[k] + carry
            carry = total >> 8
            acc = total & 0xFF
            s[k] = acc

        self.al = acc
        return acc

    def d6(self) -> int:
        """sub_18E5F：取低三位元，>=6 就重抽，回傳 1..6。"""
        while True:
            value = self.raw() & 7
            self.al = value
            if value < 6:
                return value + 1

    def roll(self, n: int) -> int:
        """sub_18E41：n<=1 原樣回傳；否則遮罩到 2^bits(n)-1，>=n 重抽，回傳 1..n。"""
        n &= 0xFF
        if n <= 1:
            return n
        mask, x = 0, n
        while x:  # stc / rcl ah,1 / shr al,1
            mask = ((mask << 1) | 1) & 0xFF
            x >>= 1
        while True:
            value = self.raw() & mask
            self.al = value
            if value < n:
                return value + 1

    def sum_d6(self, base: int, count: int) -> int:
        """sub_19D86：把 count 顆 d6 累加到 base 上。回傳 16-bit（高位溢位就飽和成 0xFFFF）。"""
        lo, hi = base & 0xFF, 0
        for _ in range(count & 0xFF):
            total = lo + self.d6()
            lo = total & 0xFF
            if total > 0xFF:
                hi = (hi + 1) & 0xFF
                if hi == 0:
                    hi, lo = 0xFF, 0xFF
        return (hi << 8) | lo

    def pair_d6(self) -> int:
        """sub_19C84：擲一對 d6 並累加，兩顆同點就再擲一對，直到不同點為止。"""
        acc = 0
        while True:
            first = self.d6()
            second = self.d6()
            acc = (acc + first + second) & 0xFF
            if second != first:
                return acc


def _self_test() -> None:
    from collections import Counter
    from math import comb

    # 全零狀態、AL=0 起步時，五重進位鏈等於五重前綴和，
    # 也就是二項式係數——直到進位開始回饋為止（第 8 項起分歧）。
    rng = Rng()
    head = [rng.raw(0) for _ in range(8)]
    assert head[:7] == [comb(n + 4, 5) % 256 for n in range(1, 8)], head
    assert head == [1, 6, 21, 56, 126, 252, 206, 25], head

    # 值域
    rng = Rng()
    assert {rng.d6() for _ in range(2000)} == set(range(1, 7))
    assert rng.roll(0) == 0 and rng.roll(1) == 1
    assert {rng.roll(3) for _ in range(2000)} == {1, 2, 3}
    assert {rng.roll(100) for _ in range(20000)} == set(range(1, 101))

    # 分佈：d6 六面各約 1/6
    rng = Rng()
    hist = Counter(rng.d6() for _ in range(600_000))
    assert all(abs(v - 100_000) < 3_000 for v in hist.values()), hist

    # 複合骰的期望值
    rng = Rng()
    mean = sum(rng.sum_d6(0, 6) for _ in range(200_000)) / 200_000
    assert abs(mean - 21.0) < 0.1, mean  # 6d6
    rng = Rng()
    mean = sum(rng.pair_d6() for _ in range(300_000)) / 300_000
    assert abs(mean - 8.4) < 0.1, mean  # 2d6 逢同點續擲

    # 狀態在三百萬次呼叫內不重複
    rng, seen = Rng(), set()
    for _ in range(3_000_000):
        rng.raw(0)
        key = tuple(rng.state)
        assert key not in seen
        seen.add(key)

    print("全部通過")


if __name__ == "__main__":
    _self_test()
