package corduroy;

import junit.framework.TestCase;

import java.util.*;

/**
 * Created by tysont on 7/15/17.
 */
public class NodeTest extends TestCase {

    public void testSend() throws Exception {

        List<Node> nodes = Helper.createCluster(2, true, false);
        Message requestMessage = new Message("Hello World!");
        Message responseMessage = nodes.get(0).send(requestMessage, nodes.get(1).getAddress());
        String text = responseMessage.getPayload();
        for(Node node : nodes) {
            node.kill();
        }

        assertEquals("HELLO WORLD!", text);
    }

    public void testProbe() throws Exception {

        List<Node> nodes = Helper.createCluster(3, true, true);
        Set<String> addresses = nodes.get(0).ping();
        for(Node node : nodes) {
            node.kill();
        }

        assertNotNull(addresses);
        assertEquals(3, addresses.size());
        assertTrue(addresses.contains(nodes.get(nodes.size() - 1).getAddress()));


    }
}