package io.projectriff.tck;

import java.util.concurrent.atomic.AtomicInteger;
import java.util.function.Function;

public class Counter implements Function<Integer, Integer> {

    private AtomicInteger total;

    @Override
    public Integer apply(Integer delta) {
        return total.getAndAdd(delta);
    }
}
