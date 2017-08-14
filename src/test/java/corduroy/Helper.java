package corduroy;

import java.util.ArrayList;
import java.util.List;

/**
 * Created by tysont on 8/14/17.
 */
public class Helper {

    public static List<Node> createCluster(int size, boolean listen, boolean discover) {

        int port = 8080;
        List<Node> nodes = new ArrayList<Node>();

        for (int i = 0; i < size; i++) {

            Node node;
            node = new Node(port++);

            if (listen) {
                Thread t = new Thread(node);
                t.start();
            }

            nodes.add(node);
        }

        if (listen) {
            try { Thread.sleep(1000); }
            catch (Exception ex) { }
        }

        if (listen && discover) {

            Node last = null;
            for (Node node : nodes) {
                if (last != null) {
                    node.discover(last.getAddress());
                }
                last = node;
            }

            nodes.get(0).discover(nodes.get(nodes.size() - 1).getAddress());
        }

        return nodes;
    }

    public static List<String> getAddresses(List<Node> nodes) {

        List<String> addresses = new ArrayList<String>();
        for (Node node : nodes) {
            addresses.add(node.getAddress());
        }

        return addresses;
    }
}
