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

        Message requestMessage = new Message("Hello World!");
        Message responseMessage = node.send(requestMessage, node.getAddress());
        String text = responseMessage.getPayload();

        node.kill();
        assertEquals("HELLO WORLD!", text);
    }

    public void testProbePayload() throws Exception {

        Node node1 = new Node(8081);
        Thread t1 = new Thread(node1);
        t1.start();

        Node node2 = new Node(8082);
        Thread t2 = new Thread(node2);
        t2.start();

        Node node3 = new Node(8083);
        Thread t3 = new Thread(node3);
        t3.start();

        try {
            Thread.sleep(1000);
        }
        catch (Exception ex) { }

        node3.initialize(node1.getAddress());
        node2.initialize(node3.getAddress());
        node1.initialize(node2.getAddress());

        Message requestMessage = new Message(new ProbePayload());
        Message responseMessage = Node.send(requestMessage, node1.getAddress());
        ProbePayload responseProbePayload = responseMessage.getPayload();
        for (String address : responseProbePayload.getAddresses()) {
            System.out.println(address);
        }
    }
}