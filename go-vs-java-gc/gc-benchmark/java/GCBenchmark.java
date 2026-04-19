import java.lang.management.GarbageCollectorMXBean;
import java.lang.management.ManagementFactory;

public class GCBenchmark {
    static byte[][] sink;

    public static void main(String[] args) {
        System.gc();
        try { Thread.sleep(100); } catch (Exception e) {}

        long gcCountBefore = 0, gcTimeBefore = 0;
        for (GarbageCollectorMXBean gc : ManagementFactory.getGarbageCollectorMXBeans()) {
            gcCountBefore += gc.getCollectionCount();
            gcTimeBefore += gc.getCollectionTime();
        }

        int liveSet = 10000;
        int allocs = 10_000_000;
        sink = new byte[liveSet][];

        long start = System.nanoTime();

        for (int i = 0; i < allocs; i++) {
            byte[] obj = new byte[64];
            obj[0] = (byte) i;
            sink[i % liveSet] = obj;
        }

        long elapsed = System.nanoTime() - start;

        long gcCountAfter = 0, gcTimeAfter = 0;
        for (GarbageCollectorMXBean gc : ManagementFactory.getGarbageCollectorMXBeans()) {
            gcCountAfter += gc.getCollectionCount();
            gcTimeAfter += gc.getCollectionTime();
            System.out.printf("  GC [%s]: count=%d, time=%dms%n",
                gc.getName(), gc.getCollectionCount(), gc.getCollectionTime());
        }

        Runtime rt = Runtime.getRuntime();
        System.out.printf("=== Java GC Benchmark ===%n");
        System.out.printf("Java version: %s%n", System.getProperty("java.version"));
        System.out.printf("VM: %s%n", System.getProperty("java.vm.name"));
        System.out.printf("Workload: %d allocs, live set %d × 64B%n", allocs, liveSet);
        System.out.printf("Total time: %.2f ms%n", elapsed / 1_000_000.0);
        System.out.printf("GC cycles: %d%n", gcCountAfter - gcCountBefore);
        System.out.printf("GC total time: %d ms%n", gcTimeAfter - gcTimeBefore);
        System.out.printf("Heap used: %.2f MB%n", (rt.totalMemory() - rt.freeMemory()) / 1024.0 / 1024.0);
        System.out.printf("Heap max: %.2f MB%n", rt.maxMemory() / 1024.0 / 1024.0);
    }
}
