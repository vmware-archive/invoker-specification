package io.projectriff.tck;

import reactor.util.function.Tuple2;
import reactor.core.publisher.Flux;

import java.util.Collections;
import java.util.function.Function;

public class Repeater implements Function<Tuple2<Flux<String>, Flux<Integer>>, Flux<String>> {

    @Override
    public Flux<String> apply(Tuple2<Flux<String>, Flux<Integer>> input) {
        Flux<String> words = input.getT1();
        Flux<Integer> numbers = input.getT2();
        return words.zipWith(numbers).flatMap(e -> Flux.fromIterable(Collections.nCopies(e.getT2(), e.getT1())));
    }
}
