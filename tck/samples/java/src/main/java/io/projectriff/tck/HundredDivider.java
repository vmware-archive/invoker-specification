package io.projectriff.tck;

import java.util.function.Function;

public class HundredDivider implements Function<Integer, Integer> {

    @Override
    public Integer apply(Integer integer) {
        return 100 / integer;
    }
}
