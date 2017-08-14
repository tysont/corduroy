package corduroy;

import junit.framework.TestCase;

import java.net.InetAddress;
import java.util.*;

/**
 * Created by tysont on 7/11/17.
 */
public class UtilityTest extends TestCase {

    public void testFindSuccessor() throws Exception {

        List<Integer> l = Arrays.asList(3, 5, 7, 9);
        assertEquals(3, Utility.findSuccessor(1, l));
        assertEquals(5, Utility.findSuccessor(4, l));
        assertEquals(5, Utility.findSuccessor(5, l));
        assertEquals(3, Utility.findSuccessor(10, l));
    }

    public void testCreateFingerTable() throws Exception {

        List<Node> nodes = Helper.createCluster(100, false, false);
        Set<String> addresses = new HashSet(Helper.getAddresses(nodes));
        Map<Integer, String> fingerTable = Utility.createFingerTable(nodes.get(0).getAddress(), addresses);
        assertNotNull(fingerTable);
        assertTrue(fingerTable.size() > 0);
    }

    public void testHashAddress() throws Exception {
        Node node = new Node(8080);
        int h1 = Utility.hashAddress(node.getAddress());
        assertTrue(h1 > 0);

        int h2 = Utility.hashAddress(node.getAddress());
        assertEquals(h1, h2);
    }

    public void testHashAddresses() throws Exception {
        Node node1 = new Node(8080);
        Node node2 = new Node(8081);

        Set<String> addresses = new HashSet<String>();
        addresses.add(node1.getAddress());
        addresses.add(node2.getAddress());

        Map<Integer, String> hashes = Utility.hashAddresses(addresses);
        assertTrue(hashes.size() == addresses.size());
    }

    public void testByteArrayToInt() throws Exception {
        assertEquals(42, Utility.byteArrayToInt(Utility.intToByteArray(42)));
    }

    public void testHash() throws Exception {
        assertNotSame(Utility.hash("foo", 1), Utility.hash("bar", 1));
        assertNotSame(Utility.hash("foo", 1), Utility.hash("foo", 2));
    }

    public void testGetAddress() throws Exception {
        assertEquals("127.0.0.1:8080", Utility.getAddress(InetAddress.getByName("127.0.0.1"), 8080));
        assertEquals("127.0.0.1:8080", Utility.getAddress(InetAddress.getByName("localhost"), 8080));
    }

    public void testGetInetAddress() throws Exception {
        assertEquals(InetAddress.getByName("127.0.0.1"), Utility.getInetAddress("127.0.0.1:8080"));
    }

    public void testGetPort() throws Exception {
        assertEquals(8080, Utility.getPort("127.0.0.1:8080"));
    }

}