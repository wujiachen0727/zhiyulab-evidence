import java.time.Duration;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.concurrent.Callable;
import java.util.concurrent.CompletionService;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorCompletionService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.concurrent.TimeUnit;

public class Main {
    record Task(String name, Duration delay, boolean fail) {}
    record Result(String name, String value) {}

    static Result runTask(Task task) throws Exception {
        Thread.sleep(task.delay().toMillis());
        if (task.fail()) {
            throw new IllegalStateException(task.name() + " returned error");
        }
        return new Result(task.name(), task.name() + "-ok");
    }

    static Summary orchestrate(String scenario, List<Task> tasks, Duration timeout) throws Exception {
        try (var executor = Executors.newVirtualThreadPerTaskExecutor()) {
            CompletionService<Result> completion = new ExecutorCompletionService<>(executor);
            List<Future<Result>> futures = new ArrayList<>();
            for (Task task : tasks) {
                Callable<Result> callable = () -> runTask(task);
                futures.add(completion.submit(callable));
            }

            Summary summary = new Summary(scenario);
            long deadline = System.nanoTime() + timeout.toNanos();
            int remaining = tasks.size();
            while (remaining > 0) {
                long waitNanos = deadline - System.nanoTime();
                if (waitNanos <= 0) {
                    summary.error = "overall timeout after " + timeout.toMillis() + "ms";
                    cancelUnfinished(futures);
                    break;
                }

                Future<Result> future = completion.poll(waitNanos, TimeUnit.NANOSECONDS);
                if (future == null) {
                    summary.error = "overall timeout after " + timeout.toMillis() + "ms";
                    cancelUnfinished(futures);
                    break;
                }

                remaining--;
                try {
                    Result result = future.get();
                    summary.completed.add(result.name());
                } catch (ExecutionException e) {
                    summary.failed.add(rootMessage(e));
                    if (summary.error == null) {
                        summary.error = rootMessage(e);
                    }
                    cancelUnfinished(futures);
                    break;
                }
            }

            for (int i = 0; i < futures.size(); i++) {
                Future<Result> future = futures.get(i);
                if (future.isCancelled()) {
                    summary.canceled.add(tasks.get(i).name());
                }
            }
            summary.sort();
            return summary;
        }
    }

    static void cancelUnfinished(List<Future<Result>> futures) {
        for (Future<Result> future : futures) {
            if (!future.isDone()) {
                future.cancel(true);
            }
        }
    }

    static String rootMessage(Exception e) {
        Throwable current = e;
        while (current.getCause() != null) {
            current = current.getCause();
        }
        return current.getMessage();
    }

    static void printSummary(Summary summary) {
        System.out.printf("scenario=%s%n", summary.scenario);
        System.out.println("state_owner=caller keeps aggregation state in ordinary objects; virtual threads do not change shared-state semantics");
        System.out.println("waiting_owner=blocking Future/CompletionService code stays synchronous; JVM virtual threads absorb most waiting cost");
        System.out.println("failure_boundary=application orchestration cancels unfinished futures after timeout or first failure");
        System.out.printf("completed=%s%n", summary.completed);
        System.out.printf("failed=%s%n", summary.failed);
        System.out.printf("canceled=%s%n", summary.canceled);
        System.out.printf("error=%s%n", summary.error == null ? "<nil>" : summary.error);
        System.out.println("---");
    }

    static class Summary {
        final String scenario;
        final List<String> completed = new ArrayList<>();
        final List<String> failed = new ArrayList<>();
        final List<String> canceled = new ArrayList<>();
        String error;

        Summary(String scenario) {
            this.scenario = scenario;
        }

        void sort() {
            completed.sort(Comparator.naturalOrder());
            failed.sort(Comparator.naturalOrder());
            canceled.sort(Comparator.naturalOrder());
        }
    }

    public static void main(String[] args) throws Exception {
        List<Task> successTasks = List.of(
            new Task("profile", Duration.ofMillis(30), false),
            new Task("billing", Duration.ofMillis(45), false),
            new Task("risk", Duration.ofMillis(20), false)
        );
        List<Task> timeoutTasks = List.of(
            new Task("profile", Duration.ofMillis(30), false),
            new Task("billing", Duration.ofMillis(45), false),
            new Task("risk", Duration.ofMillis(120), false)
        );
        List<Task> failureTasks = List.of(
            new Task("profile", Duration.ofMillis(30), false),
            new Task("billing", Duration.ofMillis(45), true),
            new Task("risk", Duration.ofMillis(80), false)
        );

        printSummary(orchestrate("java-success", successTasks, Duration.ofMillis(100)));
        printSummary(orchestrate("java-timeout", timeoutTasks, Duration.ofMillis(70)));
        printSummary(orchestrate("java-worker-error", failureTasks, Duration.ofMillis(100)));
    }
}
