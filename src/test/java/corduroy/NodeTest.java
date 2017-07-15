package corduroy;

import junit.framework.TestCase;

/**
 * Created by tysont on 7/15/17.
 */
public class NodeTest extends TestCase {

    public void testSend() throws Exception {

        Node node = new Node(8081);
        System.out.println(node.getAddress());

        Thread t = new Thread(node);
        t.start();

        try {
            Thread.sleep(1000);
        }
        catch (Exception ex) { }

        Message m1 = new Message();
        m1.setPayload("Hello World!");

        Message m2 = node.send(m1, node.getAddress());
        String text = m2.getPayload();

        node.kill();
        assertEquals("HELLO WORLD!", text);
    }
}