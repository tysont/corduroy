package corduroy;

import junit.framework.TestCase;

import java.net.InetAddress;

/**
 * Created by tysont on 7/11/17.
 */
public class UtilityTest extends TestCase {

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