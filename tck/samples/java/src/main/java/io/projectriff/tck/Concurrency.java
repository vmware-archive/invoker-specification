package io.projectriff.tck;

import java.util.function.Function;

public class Concurrency implements Function<Integer, Integer> {
    @Override
    public Integer apply(Integer integer) {
        try {
            Thread.sleep(integer);
        } catch (InterruptedException e) {
            throw new RuntimeException(e);
        }
        return integer;
    }
}
